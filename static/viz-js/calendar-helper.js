// /**
//  * Helps with the travel calendar to provide better rendering of things
//  */

// // import * as gridjs from 'gridjs';

function renderMemberName(Member) {
    const name = Member.CongressMemberInfo.name.official_full;
    const lastTerm = Member.CongressMemberInfo.terms[Member.CongressMemberInfo.terms.length - 1];

    if (lastTerm.type == "rep") {
        return `${name} (${lastTerm.party}-${lastTerm.state}-${lastTerm.district})`;
    }

    return `${name} (${lastTerm.party}-${lastTerm.state})`;
}

document.addEventListener('DOMContentLoaded', function () {
    const urlParams = new URLSearchParams(window.location.search);
    var dateFilter = urlParams.get('targetDate');

    function updateFilters(updateUrl = true) {
        // Generate the url and set the data value
        var url = `/json/travel/calendar/${yearInt}/${monthInt}`;
        if (dateFilter !== null) {
            url = '/json/travel/calendar2?' + new URLSearchParams({ targetDate: dateFilter });
            if (updateUrl) {
                const newUrl = new URL(window.location);
                newUrl.searchParams.set('targetDate', dateFilter);
                window.history.pushState({}, '', newUrl);
            }
        } else if (updateUrl) {
            const newUrl = new URL(window.location);
            newUrl.searchParams.delete('targetDate');
            window.history.pushState({}, '', newUrl);
        }

        table.setData(url);

        filtersDIV.innerHTML = '';
        if (dateFilter !== null) {
            const filter = document.createElement('div');
            filter.classList.add('inline-flex', 'items-center', 'px-3', 'py-1', 'rounded-full', 'text-sm', 'font-medium', 'bg-blue-100', 'text-blue-800', 'mr-2', 'mb-2');
            filter.innerText = "Date: " + dateFilter;
            
            var removeBtn = document.createElement('button');
            removeBtn.addEventListener('click', resetDateFilter);
            removeBtn.innerText = ' ❌';
            removeBtn.classList.add('ml-2', 'text-blue-600', 'hover:text-blue-800');

            filter.appendChild(removeBtn);
            filtersDIV.appendChild(filter);
        }
    }

    function resetDateFilter() {
        dateFilter = null;
        updateFilters();
    }

    const travelDays = document.querySelectorAll('.travelDay');
    var filtersDIV = document.getElementById('filtersApplied');

    var table = new Tabulator("#wrapper", {
        ajaxURL: dateFilter ? '/json/travel/calendar2?' + new URLSearchParams({ targetDate: dateFilter }) : `/json/travel/calendar/${yearInt}/${monthInt}`,
        pagination: true,
        // progressiveLoad:"scroll",
        paginationSize:20,
        height: '500px',
        groupBy: "Destination",
        columns: [
            {
                title: "Filer Name",
                field: 'FilerName',
            },
            {
                title: 'Congress Person',
                field: 'Member',
                formatter: (cell, formatterParams) => {
                    var value = cell.getValue();
                    if (!value) return "Unknown";
                    return `<a href="/congress-member/${value.BioGuideId}">${renderMemberName(value)}</a>`;
                 },
            },
            {
                title: 'Sponsor',
                field: 'TravelSponsor',
            },
            {
                title: 'Departure Date',
                field: 'DepartureDate',
                formatter: shittyDateFormat,
            },
            {
                title: 'Return Date',
                field: 'ReturnDate',
                formatter: shittyDateFormat,
            },
            {
                title: 'Destination',
                field: 'Destination'
            },
        ]
    });

    if (dateFilter) {
        // Need to wait a bit for Tabulator to be ready or just use updateFilters which handles it
        // Actually table.setData works fine if called immediately after constructor usually, 
        // but we already set ajaxURL in constructor.
        // Let's just run updateFilters(false) to set the filter UI.
        updateFilters(false);
    }

    travelDays.forEach((day) => {
        day.addEventListener('click', (e) => {
            // Find the td element even if a child was clicked
            const td = e.target.closest('td');
            const dayNum = td.querySelector('.day-number').innerText.trim();
            const month = td.getAttribute('data-month');
            const year = yearInt; // from global scope

            // Format as YYYY-MM-DD, ensuring 2 digits for month and day
            const formattedMonth = month.padStart(2, '0');
            const formattedDay = dayNum.padStart(2, '0');
            dateFilter = `${year}-${formattedMonth}-${formattedDay}`;
            updateFilters();
        });
    });

    // Handle back/forward buttons
    window.addEventListener('popstate', function() {
        const urlParams = new URLSearchParams(window.location.search);
        dateFilter = urlParams.get('targetDate');
        updateFilters(false);
    });
});

function shittyDateFormat(cell, formatterParams) {
    var value = cell.getValue();
    return new Date(value).toLocaleDateString();
}